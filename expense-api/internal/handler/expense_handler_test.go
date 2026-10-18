package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"expense-api/internal/handler"
	"expense-api/internal/model"
	"expense-api/internal/storage"

	"github.com/go-chi/chi/v5"
)

func setupTestStorage(t *testing.T) (*storage.Storage, func()) {
	t.Helper()
	tempDir, err := os.MkdirTemp("", "testdb_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test_expenses.db")
	store, err := storage.NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("failed to init test sqlite: %v", err)
	}

	teardown := func() {
		_ = store.Close()
		_ = os.RemoveAll(tempDir)
	}

	return store, teardown
}

func TestExpenseHandler_Create_Success(t *testing.T) {
	store, teardown := setupTestStorage(t)
	defer teardown()

	h := handler.NewExpenseHandler(store)

	payload := []byte(`{
		"amount": 24.50,
		"category": "food",
		"note": "Lunch with team",
		"spent_on": "2026-05-08"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.Expense
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.ID == 0 {
		t.Errorf("expected generated ID > 0, got %d", resp.ID)
	}
	if resp.Amount != 24.50 {
		t.Errorf("expected amount 24.50, got %f", resp.Amount)
	}
	if resp.Category != "food" {
		t.Errorf("expected category 'food', got '%s'", resp.Category)
	}
	if resp.Note != "Lunch with team" {
		t.Errorf("expected note 'Lunch with team', got '%s'", resp.Note)
	}
	if resp.SpentOn != "2026-05-08" {
		t.Errorf("expected spent_on '2026-05-08', got '%s'", resp.SpentOn)
	}
}

func TestExpenseHandler_Create_InvalidAmount(t *testing.T) {
	store, teardown := setupTestStorage(t)
	defer teardown()

	h := handler.NewExpenseHandler(store)

	testCases := []struct {
		name    string
		payload string
	}{
		{"negative amount", `{"amount": -10, "category": "food", "spent_on": "2026-05-08"}`},
		{"zero amount", `{"amount": 0, "category": "food", "spent_on": "2026-05-08"}`},
		{"too many decimal places", `{"amount": 24.555, "category": "food", "spent_on": "2026-05-08"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/expenses", bytes.NewReader([]byte(tc.payload)))
			w := httptest.NewRecorder()

			h.Create(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 Bad Request for %s, got %d", tc.name, w.Code)
			}
		})
	}
}

func TestExpenseHandler_GetAllExpenses(t *testing.T) {
	store, teardown := setupTestStorage(t)
	defer teardown()

	h := handler.NewExpenseHandler(store)

	// 1. Initial empty list returns empty JSON array []
	req := httptest.NewRequest(http.MethodGet, "/expenses", nil)
	w := httptest.NewRecorder()
	h.GetAllExpenses(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var expenses []model.Expense
	if err := json.NewDecoder(w.Body).Decode(&expenses); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(expenses) != 0 {
		t.Fatalf("expected 0 expenses, got %d", len(expenses))
	}

	// 2. Insert two expenses
	_, err := store.Create(req.Context(), model.CreateExpenseRequest{
		Amount:   15.50,
		Category: "transport",
		SpentOn:  "2026-05-01",
	})
	if err != nil {
		t.Fatalf("failed to seed expense: %v", err)
	}

	_, err = store.Create(req.Context(), model.CreateExpenseRequest{
		Amount:   45.00,
		Category: "groceries",
		SpentOn:  "2026-05-02",
	})
	if err != nil {
		t.Fatalf("failed to seed expense: %v", err)
	}

	// 3. Fetch all expenses
	w = httptest.NewRecorder()
	h.GetAllExpenses(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if err := json.NewDecoder(w.Body).Decode(&expenses); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got %d", len(expenses))
	}

	// Newest first check
	if expenses[0].Category != "groceries" {
		t.Errorf("expected first expense to be 'groceries' (newest), got '%s'", expenses[0].Category)
	}
}

func TestExpenseHandler_GetExpense(t *testing.T) {
	store, teardown := setupTestStorage(t)
	defer teardown()

	h := handler.NewExpenseHandler(store)
	r := chi.NewRouter()
	r.Get("/expenses/{id}", h.GetExpense)

	// 1. Create a sample expense
	created, err := store.Create(context.Background(), model.CreateExpenseRequest{
		Amount:   50.25,
		Category: "entertainment",
		Note:     "Cinema ticket",
		SpentOn:  "2026-05-10",
	})
	if err != nil {
		t.Fatalf("failed to create test expense: %v", err)
	}

	// 2. Fetch existing expense by ID
	req := httptest.NewRequest(http.MethodGet, "/expenses/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var fetched model.Expense
	if err := json.NewDecoder(w.Body).Decode(&fetched); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if fetched.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, fetched.ID)
	}
	if fetched.Amount != 50.25 {
		t.Errorf("expected amount 50.25, got %f", fetched.Amount)
	}
	if fetched.Note != "Cinema ticket" {
		t.Errorf("expected note 'Cinema ticket', got '%s'", fetched.Note)
	}

	// 3. Fetch non-existing expense (404)
	req404 := httptest.NewRequest(http.MethodGet, "/expenses/9999", nil)
	w404 := httptest.NewRecorder()
	r.ServeHTTP(w404, req404)

	if w404.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for non-existing id, got %d", w404.Code)
	}

	// 4. Invalid ID parameter (400)
	req400 := httptest.NewRequest(http.MethodGet, "/expenses/abc", nil)
	w400 := httptest.NewRecorder()
	r.ServeHTTP(w400, req400)

	if w400.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for invalid id, got %d", w400.Code)
	}
}

func TestExpenseHandler_DeleteExpense(t *testing.T) {
	store, teardown := setupTestStorage(t)
	defer teardown()

	h := handler.NewExpenseHandler(store)
	r := chi.NewRouter()
	r.Get("/expenses/{id}", h.GetExpense)
	r.Delete("/expenses/{id}", h.DeleteExpense)

	// 1. Create a sample expense
	_, err := store.Create(context.Background(), model.CreateExpenseRequest{
		Amount:   30.00,
		Category: "books",
		Note:     "Golang book",
		SpentOn:  "2026-05-12",
	})
	if err != nil {
		t.Fatalf("failed to create test expense: %v", err)
	}

	// 2. Delete existing expense -> 200 OK
	delReq := httptest.NewRequest(http.MethodDelete, "/expenses/1", nil)
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("expected 200 OK on delete, got %d: %s", delW.Code, delW.Body.String())
	}

	// 3. Verify it is no longer retrievable (Soft delete check -> 404)
	getReq := httptest.NewRequest(http.MethodGet, "/expenses/1", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after deletion, got %d", getW.Code)
	}

	// 4. Try deleting the already deleted expense -> 404 Not Found
	delReq2 := httptest.NewRequest(http.MethodDelete, "/expenses/1", nil)
	delW2 := httptest.NewRecorder()
	r.ServeHTTP(delW2, delReq2)

	if delW2.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when deleting already deleted expense, got %d", delW2.Code)
	}

	// 5. Try deleting non-existent ID -> 404 Not Found
	delReq404 := httptest.NewRequest(http.MethodDelete, "/expenses/9999", nil)
	delW404 := httptest.NewRecorder()
	r.ServeHTTP(delW404, delReq404)

	if delW404.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when deleting non-existent expense, got %d", delW404.Code)
	}
}
