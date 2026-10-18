package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"expense-api/internal/model"
)

var ErrNotFound = errors.New("expense not found")

type Storage struct {
	db *sql.DB
}

// Open sql lite db file if not exists create new file and create tables
func NewSQLite(dbPath string) (*Storage, error) {
	dir := filepath.Dir(dbPath)
	if dir != "." && dir != "" {
		// 0755 is best practic permission form. Owner can read, write, and execute. Group and others can read and execute but can't write.
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
	}

	// Lazy open DB file
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Really open DB file to verify the DB connection is succesfully
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create Storage instance
	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return s, nil
}

// migrate - create expenses table if not exists.
func (s *Storage) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS expenses (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount REAL NOT NULL,
		category TEXT NOT NULL,
		note TEXT,
		spent_on TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME,
		deleted_at DATETIME
	);
	`
	_, err := s.db.Exec(query)
	return err
}

// Close - close DB connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

// Create insert new expense record into SQLite database and return the created object.
func (s *Storage) Create(ctx context.Context, req model.CreateExpenseRequest) (*model.Expense, error) {
	now := time.Now().UTC()
	query := `
	INSERT INTO expenses (amount, category, note, spent_on, created_at)
	VALUES (?, ?, ?, ?, ?)
	`
	res, err := s.db.ExecContext(ctx, query, req.Amount, req.Category, req.Note, req.SpentOn, now)
	if err != nil {
		return nil, err
	}

	// return ROWID
	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &model.Expense{
		ID:        id,
		Amount:    req.Amount,
		Category:  req.Category,
		Note:      req.Note,
		SpentOn:   req.SpentOn,
		CreatedAt: now,
	}, nil
}

// GetAllExpenses returns all active (non-deleted) expenses ordered by created_at desc (newest first).
func (s *Storage) GetAllExpenses(ctx context.Context) ([]model.Expense, error) {
	query := `
	SELECT id, amount, category, COALESCE(note, ''), spent_on, created_at, updated_at
	FROM expenses
	WHERE deleted_at IS NULL
	ORDER BY created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query expenses: %w", err)
	}
	defer rows.Close()

	expenses := make([]model.Expense, 0)
	for rows.Next() {
		var exp model.Expense
		if err := rows.Scan(
			&exp.ID,
			&exp.Amount,
			&exp.Category,
			&exp.Note,
			&exp.SpentOn,
			&exp.CreatedAt,
			&exp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan expense row: %w", err)
		}
		expenses = append(expenses, exp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating expense rows: %w", err)
	}

	return expenses, nil
}

// GetExpense retrieves a single active (non-deleted) expense by its ID.
func (s *Storage) GetExpense(ctx context.Context, id int64) (*model.Expense, error) {
	query := `
	SELECT id, amount, category, COALESCE(note, ''), spent_on, created_at, updated_at
	FROM expenses
	WHERE id = ? AND deleted_at IS NULL
	`

	var exp model.Expense
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&exp.ID,
		&exp.Amount,
		&exp.Category,
		&exp.Note,
		&exp.SpentOn,
		&exp.CreatedAt,
		&exp.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get expense: %w", err)
	}

	return &exp, nil
}
