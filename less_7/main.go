package main

import (
	"log"

	"github.com/you/monorepo/ledger/internal/cache"
	"github.com/you/monorepo/ledger/internal/db"
)

func main() {
	db.Init()
	cache.Init()
	log.Println("Ledger service ready (PostgreSQL + Redis)")
	select {}
}
