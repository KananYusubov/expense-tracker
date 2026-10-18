package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"expense-api/internal/config"
	"expense-api/internal/handler"
	"expense-api/internal/logger"
	"expense-api/internal/storage"
)

func main() {
	cfg := config.Load()

	// Logger service
	appLogger, err := logger.New(cfg.LogPath)
	if err != nil {
		log.Fatalf("Logger error: %s", err)
	}
	defer appLogger.Close()

	// initialize DB
	store, err := storage.NewSQLite(cfg.DBPath)
	if err != nil {
		log.Fatalf("DB error: %s", err)
	}
	defer store.Close()

	// Handler
	expenseHandler := handler.NewExpenseHandler(store, cfg.MaxAmount)

	r := chi.NewRouter()

	// Middleware
	r.Use(appLogger.Middleware)
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong"}`))
	})

	// Expense endpoints
	r.Post("/expenses", expenseHandler.Create)
	r.Get("/expenses", expenseHandler.GetAllExpenses)
	r.Get("/expenses/summary", expenseHandler.GetAllExpensesSummary)
	r.Get("/expenses/{id:[0-9]+}", expenseHandler.GetExpense)
	r.Patch("/expenses/{id:[0-9]+}", expenseHandler.Update)
	r.Delete("/expenses/{id:[0-9]+}", expenseHandler.DeleteExpense)

	log.Printf("Server running at %s port ...  ", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server error: %s", err)
	}
}
