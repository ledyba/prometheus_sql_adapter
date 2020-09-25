package repo

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

var ErrUnknownDriver = errors.New("unknown driver")

var db *sql.DB
var log *zap.Logger

type DriverType = int

const (
	kInvalid = DriverType(iota)
	kSqlite
	kMySQL
)

var driver DriverType = kInvalid

func Open(url string) error {
	var err error
	log, _ = zap.NewProduction()
	switch {
	case strings.HasPrefix(url, "sqlite://"):
		db, err = sql.Open("sqlite3", url[9:])
		driver = kSqlite
	case strings.HasPrefix(url, "mysql://"):
		db, err = sql.Open("mysql", url[8:])
		driver = kMySQL
	default:
		return ErrUnknownDriver
	}
	if err == nil {
		log.Info("Database Opened", zap.String("driver", "mysql"), zap.String("url", url[8:]))
	} else {
		driver = kInvalid
	}
	return err
}

func Init() error {
	switch driver {
	case kSqlite:
		return sqliteInit()
	case kMySQL:
		//TODO
		return nil
	default:
		return ErrUnknownDriver
	}
}

func Close() {
	_ = log.Sync()
	_ = db.Close()
}
