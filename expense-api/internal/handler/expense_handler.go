package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"

	"expense-api/internal/model"
	"expense-api/internal/storage"
)

type ExpenseHandler struct {
	store *storage.Storage
}

// Constructor
func NewExpenseHandler(store *storage.Storage) *ExpenseHandler {
	return &ExpenseHandler{store: store}
}

// Create insert new expense record into SQLite database and return the created object.
func (h *ExpenseHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	// Task Validation Rule: The amount must be a positive number.
	// Negative or zero amounts must return a 400 with a clear error message.
	if req.Amount <= 0 {
		respondError(w, http.StatusBadRequest, "amount must be a positive number greater than 0")
		return
	}
	// Prevent more than 2 decimal places after point
	if scaled := req.Amount * 100; math.Abs(scaled-math.Round(scaled)) > 1e-9 {
		respondError(w, http.StatusBadRequest, "amount cannot have more than 2 decimal places")
		return
	}

	req.Category = strings.TrimSpace(req.Category)
	if req.Category == "" {
		respondError(w, http.StatusBadRequest, "category is required and cannot be empty")
		return
	}

	req.SpentOn = strings.TrimSpace(req.SpentOn)
	if req.SpentOn == "" {
		respondError(w, http.StatusBadRequest, "spent_on date is required (format: YYYY-MM-DD)")
		return
	}

	if _, err := time.Parse(model.DateFormat, req.SpentOn); err != nil {
		respondError(w, http.StatusBadRequest, "spent_on must be a valid date in YYYY-MM-DD format (e.g. 2026-05-08)")
		return
	}

	req.Note = strings.TrimSpace(req.Note)

	expense, err := h.store.Create(r.Context(), req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create expense: "+err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, expense)
}

// return response data as JSON format
func respondJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(data)
}

// return error as JSON format
func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
