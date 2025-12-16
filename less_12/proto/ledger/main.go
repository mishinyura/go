func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	service, closeFn, err := app.NewLedgerService(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer closeFn()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	ledgerv1.RegisterLedgerServiceServer(
		server,
		NewGRPCServer(service),
	)

	log.Printf("ledger grpc listening on :%s", port)
	log.Fatal(server.Serve(lis))
}
