package expenses

import (
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"net/http"
	"strconv"
)

func (h *Handler) UpdateExpense(c echo.Context) error {
	// Get the ID of the expense to update
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	// Parse the request body to get the updated expense data
	expense := new(Expense)
	if err := c.Bind(expense); err != nil {
		return err
	}

	// Update the expense in the database
	result, err := h.DB.Exec(`UPDATE expenses SET title = $1, amount = $2, note = $3, tags = $4 WHERE id = $5`, expense.Title, expense.Amount, expense.Note, pq.Array(expense.Tags), id)
	if err != nil {
		return err
	}

	// Check if the expense was found and updated
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return c.JSON(http.StatusOK, expense)
	}

	// Return the updated expense
	return c.JSON(http.StatusOK, expense)
}
