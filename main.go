package main

import (
	"fmt"
	"os"

	"github.com/Kaki256/multiple-word-search-backend/internal/handler"
	"github.com/Kaki256/multiple-word-search-backend/internal/migration"
	"github.com/Kaki256/multiple-word-search-backend/internal/pkg/config"
	"github.com/Kaki256/multiple-word-search-backend/internal/repository"
	"github.com/go-sql-driver/mysql"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"net/http"
)

type Word struct {
	Name       string `json:"name" db:"name"`
	Length     int    `json:"length" db:"length"`
	Sort       string `json:"sort" db:"sort"`
	IsUnvoiced bool   `json:"is_unvoiced" db:"is_unvoiced"`
	Dictionary string `json:"dictionary" db:"dictionary"`
}

const dictionaryName = "Basic"

func main() {
	e := echo.New()

	// middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.POST("/dictionary", postMakeDictionaryHandler)

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong\n")
	})

	e.GET("/search:strRegEx", wordSearchRegExHandler)

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

func wordSearchRegExHandler(c echo.Context) error {
	searchKey := c.Param("strRegEx")
	dictoinaryName := c.QueryParam("dictName")

}

func getDictionaryNameHandler(c echo.Context) error {
	dictionaryName := c.QueryParam("dictName")
}

func postMakeDictionaryHandler(c echo.Context) error {
	echo.New().GET("/dictionary/name", getDictionaryNameHandler)
	var dict []Word
	err := c.Bind(dict)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%+v", err))
	}

	conf := mysql.Config{
		User:      os.Getenv("DB_USERNAME"),
		Passwd:    os.Getenv("DB_PASSWORD"),
		Addr:      os.Getenv("DB_HOSTNAME") + ":" + os.Getenv("DB_PORT"),
		DBName:    os.Getenv("DB_DATABASE"),
		Collation: "utf8mb4_general_ci",
	}
	db, err := sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		echo.New().Logger.Fatal(err)
	}
	defer db.Close()

	// 辞書を作る
	_, err = db.Exec("INSERT INTO dictionary name VALUES ?", dictionaryName)

	return c.JSON(http.StatusOK, dict)
}
