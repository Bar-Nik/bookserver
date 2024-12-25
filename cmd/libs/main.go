package main

import (
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"server/internal/api"
	"server/internal/db"

	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	DSN      string `yaml:"dsn"`
	LogLevel int    `yaml:"log_level"`
}

func main() {
	yamlContent, err := os.ReadFile("./config.yml")
	if err != nil {
		log.Fatal(err)
	}

	var systemConfig Config
	err = yaml.Unmarshal(yamlContent, &systemConfig)
	if err != nil {
		log.Fatal(err)
	}

	log := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(systemConfig.LogLevel),
	}))

	// config := postgres.Open(systemConfig.DSN)
	// gormDB, err := gorm.Open(config, &gorm.Config{})
	// if err != nil {
	// 	panic("failed to connect database")
	// }

	// err = gormDB.AutoMigrate(&domain.Book{})
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	r := mux.NewRouter()

	migrator, err := migrate.New(
		"file://migrations",
		systemConfig.DSN)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := migrator.Up(); err != nil && err != migrate.ErrNoChange { // migrate.ErrNoChange
		fmt.Println(err)
		os.Exit(1)
	}

	rowSQLConn, err := sql.Open("postgres", systemConfig.DSN)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	repo := db.NewRepository(rowSQLConn)
	r.Use(api.Logging(log))

	ourServer := api.Server{
		Database: repo,
	}

	r.HandleFunc("/book", ourServer.GetBook).Methods(http.MethodGet)
	r.HandleFunc("/book", ourServer.AddBook).Methods(http.MethodPost)
	r.HandleFunc("/book", ourServer.DeleteBook).Methods(http.MethodDelete)
	r.HandleFunc("/book", ourServer.UpdateBook).Methods(http.MethodPut)
	r.HandleFunc("/books", ourServer.AllBooks).Methods(http.MethodGet)

	log.Warn("Server started")
	err = http.ListenAndServe("127.0.0.1:8080", r)
	if err != nil {
		log.Debug("Server failed")

	}
}
