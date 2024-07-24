package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", healthHandler)

	mux.HandleFunc("GET /setup/db", databaseSetupHandler)
	mux.HandleFunc("GET /setup/db/password", databasePasswordChangeHandler)

	// Pay account
	mux.HandleFunc("POST /account", addAccountHandler)
	mux.HandleFunc("DELETE /account", deleteAccountHandler)
	mux.HandleFunc("PUT /account", updateAccountHandler)
	mux.HandleFunc("GET /account", getAccountListHandler)

	// Pay account
	mux.HandleFunc("POST /category", addCategoryHandler)
	mux.HandleFunc("DELETE /category", deleteCategoryHandler)
	mux.HandleFunc("PUT /category", updateCategoryHandler)
	mux.HandleFunc("GET /category", getCategoryListHandler)

	// Pay record
	mux.HandleFunc("POST /record", addRecordHandler)
	mux.HandleFunc("DELETE /record", deleteRecordHandler)
	mux.HandleFunc("PUT /record", updateRecordHandler)
	mux.HandleFunc("GET /record", getRecordHandler)

	// Serve files for html
	mux.HandleFunc("GET /", handleStaticFiles)

	server := &http.Server{Addr: listenADDR, Handler: mux}
	go func() {
		fmt.Println("Server starting on " + listenADDR)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic("Server error: " + err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		fmt.Printf("Server forced to shutdown: %v", err)
	}

	if bleveIndex != nil {
		bleveIndex.Close()
	}
	if db != nil {
		db.Close()
	}

	fmt.Println("Server exited")
}
