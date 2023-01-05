//go:build unit
// +build unit

package expenses

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/labstack/echo/v4"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

var Data = `{
	"id":1,
	"title": "test-title",
	"amount": 45.45,
	"note": "test-note",
	"tags": ["test-tags1","test-tags2"]
}`

func TestGetExpenses(t *testing.T) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", strings.NewReader(""))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	newsMockRows := sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
		AddRow("1", "test-title", "45.45", "test-note", `{"test-tags1","test-tags2"}`)

	db, mock, err := sqlmock.New()
	mock.ExpectQuery("SELECT (.+) FROM expenses").WillReturnRows(newsMockRows)
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	h := Handler{db}
	c := e.NewContext(req, rec)
	expected := "[{\"id\":1,\"title\":\"test-title\",\"amount\":45.45,\"note\":\"test-note\",\"tags\":[\"test-tags1\",\"test-tags2\"]}]"

	err = h.GetExpenses(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expected, strings.TrimSpace(rec.Body.String()))
	}
}
func TestGetOneExpenses(t *testing.T) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	newsMockRows := sqlmock.NewRows([]string{"id", "title", "amount", "note", "tags"}).
		AddRow("1", "test-title", "45.45", "test-note", `{"test-tags1","test-tags2"}`)

	db, mock, err := sqlmock.New()
	mock.ExpectPrepare(regexp.QuoteMeta("SELECT id, title, amount, note, tags FROM expenses WHERE id = $1")).ExpectQuery().WithArgs("1").WillReturnRows(newsMockRows)

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	h := Handler{db}
	c := e.NewContext(req, rec)
	c.SetPath("/expenses/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")
	expected := "{\"id\":1,\"title\":\"test-title\",\"amount\":45.45,\"note\":\"test-note\",\"tags\":[\"test-tags1\",\"test-tags2\"]}"

	err = h.GetExpense(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expected, strings.TrimSpace(rec.Body.String()))
	}
}

func TestCreateExpenses(t *testing.T) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(Data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	newsMockRows := sqlmock.NewRows([]string{"id"}).
		AddRow("1")

	db, mock, err := sqlmock.New()

	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO expenses (title, amount, note, tags) VALUES ($1, $2, $3, $4) RETURNING id")).WithArgs("test-title", 45.45, "test-note", pq.Array([]string{"test-tags1", "test-tags2"})).WillReturnRows(newsMockRows)

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	h := Handler{db}
	c := e.NewContext(req, rec)

	expected := "{\"id\":1,\"title\":\"test-title\",\"amount\":45.45,\"note\":\"test-note\",\"tags\":[\"test-tags1\",\"test-tags2\"]}"

	h.CreateExpense(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, expected, strings.TrimSpace(rec.Body.String()))
	}
}

func TestUpdateExpense(t *testing.T) {

	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/", strings.NewReader(Data))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	db, mock, err := sqlmock.New()

	mock.ExpectExec(regexp.QuoteMeta("UPDATE expenses SET title = $1, amount = $2, note = $3, tags = $4 WHERE id = $5")).
		WithArgs("test-title", 45.45, "test-note", pq.Array([]string{"test-tags1", "test-tags2"}), 1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	h := Handler{db}
	c := e.NewContext(req, rec)
	c.SetPath("/:id")
	c.SetParamNames("id")
	c.SetParamValues("1")

	expected := "{\"id\":1,\"title\":\"test-title\",\"amount\":45.45,\"note\":\"test-note\",\"tags\":[\"test-tags1\",\"test-tags2\"]}"

	err = h.UpdateExpense(c)

	if assert.NoError(t, err) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, expected, strings.TrimSpace(rec.Body.String()))
	}
}
