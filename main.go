package main

import (
	"fmt"

	"github.com/Kaki256/multiple-word-search-backend/internal/handler"
	"github.com/Kaki256/multiple-word-search-backend/internal/migration"
	"github.com/Kaki256/multiple-word-search-backend/internal/pkg/config"
	"github.com/Kaki256/multiple-word-search-backend/internal/repository"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"net/http"
)

type Word struct {
	Name string `json:"name"`
}

type Dictionary struct {
	Words []Word `json:"words"`
}

func main() {
	e := echo.New()

	// middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.POST("/api/dictionary", postDictionaryHandler)

	// connect to database
	db, err := sqlx.Connect("mysql", config.MySQL().FormatDSN())
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	// migrate tables
	if err := migration.MigrateTables(db.DB); err != nil {
		e.Logger.Fatal(err)
	}

	// setup repository
	repo := repository.New(db)

	// setup routes
	h := handler.New(repo)
	v1API := e.Group("/api/v1")
	h.SetupRoutes(v1API)

	e.Logger.Fatal(e.Start(config.AppAddr()))
}

func postDictionaryHandler(c echo.Context) error {
	dict := &Dictionary{}
	err := c.Bind(dict)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%+v", err))
	}
	return c.JSON(http.StatusOK, dict)
}
