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
		log.Fatalf("Failed to connect to Auth: %v", err)
	}
	authClient = pb_auth.NewAuthServiceClient(connAuth)

	connLedger, err := grpc.Dial("ledger-service:50052", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Ledger: %v", err)
	}
	ledgerClient = pb_ledger.NewLedgerServiceClient(connLedger)

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/transaction", transactionHandler)
	http.HandleFunc("/report", reportHandler)

	log.Println("Gateway running on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb_auth.AuthRequest
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
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req pb_auth.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp, err := authClient.Login(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func transactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

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

	req := &pb_ledger.ReportRequest{UserId: valResp.UserId}
	resp, err := ledgerClient.GetReport(context.Background(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
