package model

import (
	"time"
)

const DateFormat = "2006-01-02"

type Expense struct {
	ID        int64      `json:"id"`
	Amount    float64    `json:"amount"`
	Category  string     `json:"category"`
	Note      string     `json:"note,omitempty"`
	SpentOn   string     `json:"spent_on"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"-"`
}
type CreateExpenseRequest struct {
	Amount   float64 `json:"amount"`
	Category string  `json:"category"`
	Note     string  `json:"note"`
	SpentOn  string  `json:"spent_on"`
}

type UpdateExpenseRequest struct {
	Amount   *float64 `json:"amount,omitempty"`
	Category *string  `json:"category,omitempty"`
	Note     *string  `json:"note,omitempty"`
	SpentOn  *string  `json:"spent_on,omitempty"`
}

type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

type ExpenseSummary struct {
	TotalAmount     float64           `json:"total_amount"`
	TotalExpenses   int64             `json:"total_expenses"`
	TotalCategories int               `json:"total_categories"`
	Categories      []CategorySummary `json:"categories"`
}
