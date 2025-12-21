package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	pb_auth "github.com/yuramishin/expense-tracker/proto/pb_auth"
	pb_ledger "github.com/yuramishin/expense-tracker/proto/pb_ledger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var authClient pb_auth.AuthServiceClient
var ledgerClient pb_ledger.LedgerServiceClient

func main() {
	connAuth, err := grpc.Dial("auth-service:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect to Auth Service: %v", err)
	}
	authClient = pb_auth.NewAuthServiceClient(connAuth)

	connLedger, err := grpc.Dial("ledger-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Did not connect to Ledger Service: %v", err)
	}
	ledgerClient = pb_ledger.NewLedgerServiceClient(connLedger)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/transaction", transactionHandler)
	http.HandleFunc("/report", reportHandler)
	http.HandleFunc("/set_budget", setBudgetHandler)
	http.HandleFunc("/get_budgets", getBudgetsHandler)

	log.Println("Gateway running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	var req pb_auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := authClient.Register(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req pb_auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := authClient.Login(context.Background(), &req)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	_, err := authClient.Logout(context.Background(), &pb_auth.LogoutRequest{Token: token})

	if err != nil {
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true, "message": "Logged out successfully"}`))
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	valResp, err := authClient.Validate(context.Background(), &pb_auth.ValidateRequest{Token: token})
	if err != nil || !valResp.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req pb_ledger.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.UserId = valResp.UserId

	resp, err := ledgerClient.CreateTransaction(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func reportHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	valResp, err := authClient.Validate(context.Background(), &pb_auth.ValidateRequest{Token: token})
	if err != nil || !valResp.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := ledgerClient.GetReport(context.Background(), &pb_ledger.ReportRequest{UserId: valResp.UserId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func setBudgetHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	valResp, err := authClient.Validate(context.Background(), &pb_auth.ValidateRequest{Token: token})
	if err != nil || !valResp.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req pb_ledger.BudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.UserId = valResp.UserId

	resp, err := ledgerClient.SetBudget(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func getBudgetsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	valResp, err := authClient.Validate(context.Background(), &pb_auth.ValidateRequest{Token: token})
	if err != nil || !valResp.Valid {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := ledgerClient.GetBudgets(context.Background(), &pb_ledger.GetBudgetsRequest{UserId: valResp.UserId})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
