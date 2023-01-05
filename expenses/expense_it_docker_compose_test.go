//go:build integration

package expenses

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

const serverPort = 80

func NewApplication(db *sql.DB) *Handler {
	return &Handler{db}
}

func TestGetExpenseByID(t *testing.T) {
	eh := echo.New()
	go func(e *echo.Echo) {
		db, err := sql.Open("postgres", "postgresql://root:root@db/expenseDB?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
		h := NewApplication(db)

		e.GET("/expenses/:id", h.GetExpense)
		e.Start(fmt.Sprintf(":%d", serverPort))
	}(eh)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", serverPort), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	expenseID := 1
	resp := request(http.MethodGet, uri("expenses", strconv.Itoa(expenseID)), nil)
	defer resp.Body.Close()

	var expense Expense
	err := resp.Decode(&expense)

	expensesBytes, err := json.Marshal(expense)
	expensesString := string(expensesBytes)

	expected := "{\"id\":1,\"title\":\"coke\",\"amount\":10.5,\"note\":\"test note\",\"tags\":[\"test-tags\"]}"

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, expected, strings.TrimSpace(string(expensesString)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = eh.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestGetAllExpenses(t *testing.T) {

	// Setup server
	eh := echo.New()
	go func(e *echo.Echo) {
		db, err := sql.Open("postgres", "postgresql://root:root@db/expenseDB?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}

		h := NewApplication(db)

		e.GET("/expenses", h.GetExpenses)
		e.Start(fmt.Sprintf(":%d", serverPort))
	}(eh)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", serverPort), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	resp := request(http.MethodGet, uri("expenses"), nil)
	defer resp.Body.Close()

	// Decode the response body into a slice of expenses
	var expenses []Expense
	err := resp.Decode(&expenses)

	expensesBytes, err := json.Marshal(expenses)
	expensesString := string(expensesBytes)

	expected := "[{\"id\":1,\"title\":\"coke\",\"amount\":10.5,\"note\":\"test note\",\"tags\":[\"test-tags\"]}]"

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, expected, strings.TrimSpace(string(expensesString)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = eh.Shutdown(ctx)
	assert.NoError(t, err)

}

func TestCreateExpense(t *testing.T) {

	eh := echo.New()
	go func(e *echo.Echo) {
		db, err := sql.Open("postgres", "postgresql://root:root@db/expenseDB?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
		h := NewApplication(db)

		e.POST("/expenses", h.CreateExpense)
		e.Start(fmt.Sprintf(":%d", serverPort))
	}(eh)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", serverPort), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	body := bytes.NewBufferString(`{
		"title": "strawberry smoothie",
		"amount": 79.45,
		"note": "night market promotion discount 10 bath",
		"tags": ["food", "beverage"]
	}`)

	var expense Expense

	res := request(http.MethodPost, uri("expenses"), body)
	err := res.Decode(&expense)
	defer res.Body.Close()

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, res.StatusCode)
	assert.NotEqual(t, 1, expense.ID)
	assert.Equal(t, "strawberry smoothie", expense.Title)
	assert.Equal(t, 79.45, expense.Amount)
	assert.Equal(t, "night market promotion discount 10 bath", expense.Note)
	assert.Equal(t, []string{"food", "beverage"}, expense.Tags)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = eh.Shutdown(ctx)
	assert.NoError(t, err)
}

func TestUpdateExpense(t *testing.T) {
	// Set up test server and send create expense request
	eh := echo.New()
	go func(e *echo.Echo) {
		db, err := sql.Open("postgres", "postgresql://root:root@db/expenseDB?sslmode=disable")
		if err != nil {
			log.Fatal(err)
		}
		h := NewApplication(db)

		e.POST("/expenses", h.CreateExpense)
		e.PUT("/expenses/:id", h.UpdateExpense)
		e.Start(fmt.Sprintf(":%d", serverPort))
	}(eh)
	for {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", serverPort), 30*time.Second)
		if err != nil {
			log.Println(err)
		}
		if conn != nil {
			conn.Close()
			break
		}
	}

	createBody := bytes.NewBufferString(`{
		"title": "strawberry smoothie",
		"amount": 79.45,
		"note": "promotion discount 10 bath",
		"tags": ["food", "beverage"]
	}`)
	var createExpense Expense
	createRes := request(http.MethodPost, uri("expenses"), createBody)
	err := createRes.Decode(&createExpense)
	defer createRes.Body.Close()
	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, createRes.StatusCode)

	// Send update expense request
	updateBody := bytes.NewBufferString(`{
		"title": "chocolate milkshake",
		"amount": 69.45,
		"note": "price 69.45",
		"tags": ["food", "beverage", "dessert"]
	}`)
	updateURL := uri("expenses", strconv.Itoa(createExpense.ID))
	var updateExpense Expense
	updateRes := request(http.MethodPut, updateURL, updateBody)
	err = updateRes.Decode(&updateExpense)
	defer updateRes.Body.Close()

	// Check response
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, updateRes.StatusCode)
	assert.Equal(t, "chocolate milkshake", updateExpense.Title)
	assert.Equal(t, 69.45, updateExpense.Amount)
	assert.Equal(t, "price 69.45", updateExpense.Note)
	assert.Equal(t, []string{"food", "beverage", "dessert"}, updateExpense.Tags)

	// Shut down server
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = eh.Shutdown(ctx)
	assert.NoError(t, err)
}

func uri(paths ...string) string {
	host := fmt.Sprintf("http://localhost:%d", serverPort)
	if paths == nil {
		return host
	}

	url := append([]string{host}, paths...)
	return strings.Join(url, "/")
}

func request(method, url string, body io.Reader) *Response {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Add("Content-Type", "application/json")
	client := http.Client{}
	res, err := client.Do(req)

	return &Response{res, err}
}

type Response struct {
	*http.Response
	err error
}

func (r *Response) Decode(v interface{}) error {
	if r.err != nil {
		return r.err
	}

	return json.NewDecoder(r.Body).Decode(v)
}
