package main

import (
	"log"
	"net/http"

	"github.com/you/monorepo/gateway/internal/api"
	"github.com/you/monorepo/ledger"
)

func main() {
	l := ledger.NewLedger()

	mux := http.NewServeMux()

	h := &api.Handlers{Ledger: l}
	h.Register(mux)

	addr := ":8080"
	log.Printf("Gateway listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
