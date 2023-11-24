package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"

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
	IsUnvoiced int    `json:"is_unvoiced" db:"is_unvoiced"`
	Dictionary string `json:"dictionary" db:"dictionary"`
}

var db *sqlx.DB

func main() {
	var conf = mysql.Config{
		User:      os.Getenv("DB_USERNAME"),
		Passwd:    os.Getenv("DB_PASSWORD"),
		Addr:      os.Getenv("DB_HOSTNAME") + ":" + os.Getenv("DB_PORT"),
		DBName:    os.Getenv("DB_DATABASE"),
		Collation: "utf8mb4_general_ci",
	}
	var err error
	db, err = sqlx.Open("mysql", conf.FormatDSN())
	if err != nil {
		echo.New().Logger.Fatal(err)
	}
	defer db.Close()

	e := echo.New()

	// middlewares
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.CORS())

	e.POST("/dictionary", postMakeDictionaryHandler)

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong\n")
	})

	e.GET("/search/:dict", wordSearchRegExHandler)

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
	dictionaryName := c.Param("dict")
	searchKey := c.QueryParam("strRegEx")

	var dict []Word
	err := db.Get(&dict, "SELECT * FROM words WHERE dictionary = ?", dictionaryName)
	if errors.Is(err, sql.ErrNoRows) {
		log.Printf("no such dictionary name = '%s'\n", dictionaryName)
	}

	var result []Word
	regEx, err := regexp.Compile(searchKey)
	if err != nil {
		log.Printf(":(\n")
	}

	for _, word := range dict {
		if regEx.MatchString(word.Name) {
			result = append(result, word)
		}
	}

	return c.JSON(http.StatusOK, result)
}

func postMakeDictionaryHandler(c echo.Context) error {
	dictionaryName := c.QueryParam("dict")
	var dict []Word
	err := c.Bind(dict)

	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("%+v", err))
	}

	// 辞書を作る
	_, err = db.Exec("INSERT INTO dictionary name VALUES ?", dictionaryName)
	if err != nil {
		log.Fatal(err)
	}

	insert := "INSERT INTO words(name, length, sort, is_unvoiced, dictionary) VALUES "

	for _, word := range dict {
		insert += fmt.Sprintf("('%s', %d, '%s', %d, '%s')", word.Name, word.Length, word.Sort, word.IsUnvoiced, dictionaryName)
	}

	_, err = db.Exec(insert)
	if err != nil {
		log.Fatal(err)
	}

	return c.JSON(http.StatusOK, dict)
}
