package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ledyba/prometheus_sql_adapter/internal/repo"
	"github.com/ledyba/prometheus_sql_adapter/internal/web"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mattn/go-isatty"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)


var standardLog bool
var webListenAddress string
var dbUrl string

var log *zap.Logger

func main() {
	{ // Setup flags
		flag.StringVar(&webListenAddress, "web.listen-address", ":8080", "Listen")
		flag.StringVar(&dbUrl, "db.url", "sqlite://file::memory:?cache=shared", "DB to connect")
		flag.BoolVar(&standardLog, "standard-log", false, "Show log human readably")
		flag.Parse()
	}
	var err error
	// Check weather terminal or not
	if standardLog || isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
		log, err = zap.NewDevelopment()
	} else {
		log, err = zap.NewProduction()
	}
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Failed to create logger: %v", err)
		os.Exit(-1)
	}
	undo := zap.ReplaceGlobals(log)
	defer func() {
		undo()
		_ = log.Sync()
	}()
	log.Info("Log System Initialized.")

	err = repo.Open(dbUrl)
	if err != nil {
		log.Fatal("Failed to open database", zap.String("url", dbUrl), zap.Error(err))
	}
	err = repo.Init()
	if err != nil {
		log.Fatal("Failed to init database", zap.String("url", dbUrl), zap.Error(err))
	}
	defer repo.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/read", web.Read)
	mux.HandleFunc("/write", web.Write)
	log.Info("Listening", zap.String("at", webListenAddress))
	err = http.ListenAndServe(webListenAddress, mux)
	if err != nil {
		log.Fatal("Failed to run server", zap.Error(err))
	}
}
