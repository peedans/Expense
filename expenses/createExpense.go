package expenses

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"net/http"
)

func (h *Handler) CreateExpense(c echo.Context) error {
	// Parse the request body to get the expense data
	expense := Expense{}
	if err := c.Bind(&expense); err != nil {
		return err
	}

	sqlCreateExpense := `INSERT INTO expenses (title, amount, note, tags)
	VALUES ($1, $2, $3, $4)
	RETURNING id`

	// Insert the expense into the database
	err := h.DB.QueryRow(
		sqlCreateExpense, expense.Title, expense.Amount, expense.Note, pq.Array(&expense.Tags),
	).Scan(&expense.ID)
	if err != nil {
		fmt.Println(err)
	}

	return c.JSON(http.StatusCreated, expense)
}
