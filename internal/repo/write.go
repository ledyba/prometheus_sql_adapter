package repo

import (
	"github.com/prometheus/prometheus/prompb"
)

func Write(req *prompb.WriteRequest) error {
	switch driver {
	case kSqlite:
		return sqliteWrite(req)
	case kMySQL:
		return mysqlWrite(req)
	default:
		return ErrUnknownDriver
	}
}
