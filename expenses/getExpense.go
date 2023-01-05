package expenses

import (
	"database/sql"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"net/http"
)

func (h *Handler) GetExpenses(c echo.Context) error {
	rows, err := h.DB.Query("SELECT * FROM expenses")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't query all expenses:" + err.Error()})
	}

	// Loop through the rows and append each expense to a slice
	expenses := make([]Expense, 0)
	for rows.Next() {
		var expense Expense
		err := rows.Scan(&expense.ID, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses:" + err.Error()})
		}
		expenses = append(expenses, expense)
	}

	// Return the slice of expenses
	return c.JSON(http.StatusOK, expenses)
}

func (h *Handler) GetExpense(c echo.Context) error {
	id := c.Param("id")
	stmt, err := h.DB.Prepare(`SELECT id, title, amount, note, tags FROM expenses WHERE id = $1`)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query expense statement:" + err.Error()})
	}

	row := stmt.QueryRow(id)
	expense := Expense{}
	err = row.Scan(&expense.ID, &expense.Title, &expense.Amount, &expense.Note, pq.Array(&expense.Tags))
	switch err {
	case sql.ErrNoRows:
		return c.JSON(http.StatusNotFound, Err{Message: "expense not found"})
	case nil:
		return c.JSON(http.StatusOK, expense)
	default:
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expense:" + err.Error()})
	}
}
