package repo

import (
	"database/sql"
	"errors"
	"io"
	"strings"

	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

// ErrUnknownDriver indicates that given sql driver is not supported.
var ErrUnknownDriver = errors.New("unknown driver")

var db *sql.DB
var driver Driver

// Driver abstract underlying database.
type Driver interface {
	Init() error
	Read(req *prompb.ReadRequest, w io.Writer) error
	Write(req *prompb.WriteRequest) error
	Close()
}

// Open database
func Open(url string) error {
	log := zap.L()
	var err error
	log, _ = zap.NewProduction()
	switch {
	case strings.HasPrefix(url, "sqlite://"):
		db, err = sql.Open("sqlite3", url[9:])
		if err != nil {
			return err
		}
		driver = newSqlite(db)
		log.Info("Database Opened", zap.String("driver", "sqlite"), zap.String("url", url[9:]))
		return nil
	case strings.HasPrefix(url, "mysql://"):
		db, err = sql.Open("mysql", url[8:])
		if err != nil {
			return err
		}
		driver = newMysql(db)
		log.Info("Database Opened", zap.String("driver", "mysql"), zap.String("url", url[8:]))
		return nil
	}
	return ErrUnknownDriver
}

// Init drivers
func Init() error {
	if driver == nil {
		return ErrUnknownDriver
	}
	return driver.Init()
}

// Read handles read request
func Read(req *prompb.ReadRequest, w io.Writer) error {
	if driver == nil {
		return ErrUnknownDriver
	}
	return driver.Read(req, w)
}

// Write handles write request
func Write(req *prompb.WriteRequest) error {
	if driver == nil {
		return ErrUnknownDriver
	}
	return driver.Write(req)
}

// Close all resources
func Close() {
	driver.Close()
	_ = db.Close()
}
