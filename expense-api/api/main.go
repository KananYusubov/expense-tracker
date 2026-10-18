package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"expense-api/internal/handler"
	"expense-api/internal/storage"
)

func main() {
	port := os.Getenv("PORT") //TODO: make env process modular and dynamic
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./data/expenses.db"
	}

	// initialize DB
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		log.Fatalf("DB error: %s", err)
	}
	defer store.Close()

	// Handler
	expenseHandler := handler.NewExpenseHandler(store)

	r := chi.NewRouter()

	// Chi-nin hazır middleware-ləri
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong"}`))
	})

	// Expense endpoints
	r.Post("/expenses", expenseHandler.Create)
	r.Get("/expenses", expenseHandler.GetAllExpenses)
	r.Get("/expenses/{id}", expenseHandler.GetExpense)

	log.Printf("Server running at %s port ...  ", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server error: %s", err)
	}
}
