package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func SetupServer() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /setup/db", setupDatabaseHandler)

	// Pay method
	mux.HandleFunc("POST /method", addMethodHandler)
	mux.HandleFunc("DELETE /method", deleteMethodHandler)
	mux.HandleFunc("PUT /method", updateMethodHandler)
	mux.HandleFunc("GET /method", getMethodListHandler)

	// Pay record
	mux.HandleFunc("POST /record", addRecordHandler)
	mux.HandleFunc("DELETE /record", deleteRecordHandler)
	mux.HandleFunc("PUT /record", updateRecordHandler)
	mux.HandleFunc("GET /record", getRecordHandler)
	mux.HandleFunc("GET /record/sum", getSumHandler)

	mux.HandleFunc("GET /", handleStaticFiles)

	server := &http.Server{Addr: "127.0.0.1:8080", Handler: mux}
	go func() {
		fmt.Println("Server starting on localhost:8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	if bleveIndex != nil {
		bleveIndex.Close()
	}
	if db != nil {
		db.Close()
	}

	fmt.Println("Server exited")
}
