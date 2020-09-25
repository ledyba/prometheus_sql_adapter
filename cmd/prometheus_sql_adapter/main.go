package main

import (
	"flag"
	"net/http"

	"github.com/ledyba/prometheus_sql_adapter/internal/repo"
	"github.com/ledyba/prometheus_sql_adapter/internal/web"

	"go.uber.org/zap"
)

var listen = flag.String("listen", ":8080", "Listen")
var db = flag.String("db", "sqlite://file::memory:?cache=shared", "DB to connect")

var log *zap.Logger

func main() {
	var err error
	log, _ = zap.NewProduction()
	defer func() {
		_ = log.Sync()
	}()
	err = repo.Open(*db)
	if err != nil {
		log.Fatal("Failed to open database", zap.String("url", *db), zap.Error(err))
	}
	err = repo.Init()
	if err != nil {
		log.Fatal("Failed to init database", zap.String("url", *db), zap.Error(err))
	}
	defer repo.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/read", web.Read)
	mux.HandleFunc("/write", web.Write)
	log.Info("Listening", zap.String("at", *listen))
	err = http.ListenAndServe(*listen, mux)
	if err != nil {
		log.Fatal("Failed to run server", zap.Error(err))
	}
}
