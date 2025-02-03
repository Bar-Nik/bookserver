package main

import (
	pb "bookserver_git/api/proto/v1"
	"bookserver_git/internal/api"
	"bookserver_git/internal/db"
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	grpc_run "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gopkg.in/yaml.v3"

	_ "github.com/lib/pq"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type Config struct {
	DSN      string `yaml:"dsn"`
	LogLevel int    `yaml:"log_level"`
	Host     string `yaml:"host"`
	HostGRPC string `yaml:"host_grpc"`
}

func main() {

	loggingOpts := []logging.Option{
		logging.WithLogOnEvents(
			logging.StartCall,
			logging.FinishCall,
			logging.PayloadReceived,
			logging.PayloadSent,
		),
	}

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

	// r := mux.NewRouter()

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
	// r.Use(api.Logging(log))

	ourServer := api.Server{
		Database: repo,
	}

	ln, err := net.Listen("tcp", systemConfig.HostGRPC)
	if err != nil {
		fmt.Println(err)
	}
	server := grpc.NewServer(
		grpc.Creds(insecure.NewCredentials()),
		grpc.ChainUnaryInterceptor(
			// Order matters e.g. tracing interceptor have to create span first for the later exemplars to work.
			logging.UnaryServerInterceptor(interceptorLogger(log), loggingOpts...),
		),
	)
	pb.RegisterBookAPIServer(server, &ourServer)

	go func() {
		if err = server.Serve(ln); err != nil {
			fmt.Println(err)
		}
	}()

	conn, err := grpc.NewClient(systemConfig.HostGRPC,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		fmt.Println(err)
	}
	defer conn.Close()

	gw := grpc_run.NewServeMux()

	err = pb.RegisterBookAPIHandler(context.TODO(), gw, conn)
	if err != nil {
		fmt.Println(err)
	}
	gwServer := &http.Server{
		Addr:    systemConfig.Host,
		Handler: gw,
	}

	// r.HandleFunc("/book", ourServer.GetBook).Methods(http.MethodGet)
	// r.HandleFunc("/book", ourServer.AddBook).Methods(http.MethodPost)
	// r.HandleFunc("/book", ourServer.DeleteBook).Methods(http.MethodDelete)
	// r.HandleFunc("/book", ourServer.UpdateBook).Methods(http.MethodPut)
	// r.HandleFunc("/books", ourServer.AllBooks).Methods(http.MethodGet)

	log.Warn("Server started")
	err = gwServer.ListenAndServe()
	if err != nil {
		log.Debug("Server failed")
	}
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
