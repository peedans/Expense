package main

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/peedans/assessment/expenses"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewApplication(db *sql.DB) *expenses.Handler {
	return expenses.InitDB(db)
}

var db *sql.DB

func main() {
	var err error
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Println("Could not load .env file")
		os.Exit(1)
	}
	db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	defer db.Close()

	h := NewApplication(db)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == "admin" && password == "password" {
			return true, nil
		}
		return false, echo.ErrUnauthorized
	}))

	e.POST("/expenses", h.CreateExpense)
	e.PUT("/expenses/:id", h.UpdateExpense)
	e.GET("/expenses", h.GetExpenses)
	e.GET("/expenses/:id", h.GetExpense)

	// Start the web server
	port := os.Getenv("PORT")
	if port == "" {
		port = ":2565"
	}

	// Start the web server in a separate goroutine
	go func() {
		if err := e.Start(port); err != nil && err != http.ErrServerClosed {
			e.Logger.Info("shutting down the server")
		}
	}()
	fmt.Println("server starting at :2565")

	gracefulShutdown := make(chan os.Signal, 1)

	signal.Notify(gracefulShutdown, os.Interrupt, syscall.SIGTERM)

	<-gracefulShutdown
	fmt.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}
