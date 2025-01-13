package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-playground/validator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	"github.com/bxcodec/go-clean-arch/internal/helper"
	"github.com/bxcodec/go-clean-arch/internal/repository"
	mysqlRepo "github.com/bxcodec/go-clean-arch/internal/repository/mysql"
	"github.com/bxcodec/go-clean-arch/pdf"

	"github.com/bxcodec/go-clean-arch/article"
	_ "github.com/bxcodec/go-clean-arch/docs"
	"github.com/bxcodec/go-clean-arch/internal/rest"
	"github.com/bxcodec/go-clean-arch/internal/rest/middleware"
	"github.com/joho/godotenv"
)

const (
	defaultTimeout = 30
	defaultAddress = ":9090"
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// @title Swagger Example API
// @version 1.0
// @description This is a sample server for using Swagger with Echo.
// @host localhost:9090
// @BasePath /
func main() {
	//prepare database
	dbHost := os.Getenv("DATABASE_HOST")
	dbPort := os.Getenv("DATABASE_PORT")
	dbUser := os.Getenv("DATABASE_USER")
	dbPass := os.Getenv("DATABASE_PASS")
	dbName := os.Getenv("DATABASE_NAME")
	connection := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	val := url.Values{}
	val.Add("parseTime", "1")
	val.Add("loc", "Asia/Jakarta")
	dsn := fmt.Sprintf("%s?%s", connection, val.Encode())
	dbConn, err := sql.Open(`mysql`, dsn)
	if err != nil {
		log.Fatal("failed to open connection to database", err)
	}
	err = dbConn.Ping()
	if err != nil {
		log.Fatal("failed to ping database ", err)
	}

	defer func() {
		err := dbConn.Close()
		if err != nil {
			log.Fatal("got error when closing the DB connection", err)
		}
	}()
	// prepare echo

	e := echo.New()

	v := validator.New()
	e.Validator = &helper.CustomValidator{Validator: v}

	e.Use(middleware.CORS)
	timeoutStr := os.Getenv("CONTEXT_TIMEOUT")
	timeout, err := strconv.Atoi(timeoutStr)
	if err != nil {
		log.Println("failed to parse timeout, using default timeout")
		timeout = defaultTimeout
	}
	timeoutContext := time.Duration(timeout) * time.Second
	e.Use(middleware.SetRequestContextWithTimeout(timeoutContext))

	// Prepare Repository
	authorRepo := mysqlRepo.NewAuthorRepository(dbConn)
	articleRepo := mysqlRepo.NewArticleRepository(dbConn)

	pdfApi := &repository.PdfCpuApiImpl{}
	fileHelper := &repository.FileHelperImpl{}
	pdfRepo := repository.NewPdfRepository(pdfApi, fileHelper)

	// Build service Layer
	svc := article.NewService(articleRepo, authorRepo)
	rest.NewArticleHandler(e, svc)

	pdfSvc := pdf.NewService(pdfRepo)
	rest.NewPdfHandler(e, pdfSvc)

	// Swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Start Server
	address := os.Getenv("SERVER_ADDRESS")
	if address == "" {
		address = defaultAddress
	}
	log.Fatal(e.Start(address)) //nolint
}
