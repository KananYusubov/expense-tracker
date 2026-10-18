package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"expense-api/internal/handler"
	"expense-api/internal/model"
	"expense-api/internal/storage"
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
