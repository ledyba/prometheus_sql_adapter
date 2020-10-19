package repo

import (
	"database/sql"
	"errors"
	"io"
	"sync"

	"github.com/cornelk/hashmap"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

type sqliteDriver struct {
	literalCache hashmap.HashMap
	db           *sql.DB
	waitGroup    sync.WaitGroup
	writeQueue   chan<- *prompb.WriteRequest
}

func newSqlite(db *sql.DB) Driver {
	return &sqliteDriver{
		literalCache: hashmap.HashMap{},
		db:           db,
	}
}

func (d *sqliteDriver) Init() error {
	log := zap.L()

	db := d.db
	db.SetMaxOpenConns(16)
	_, err := db.Exec(`
create table if not exists timeseries(
  id integer primary key autoincrement
);
create table if not exists labels(
  timeseries_id integer not null,
  name integer not null,
  value integer not null
);
create table if not exists literals(
  id integer primary key autoincrement,
  value text unique not null
);
create table if not exists samples(
  timeseries_id integer not null,
  timestamp integer not null,
  value real not null
);

-- labels
create index if not exists labels_timeseries_index on labels(timeseries_id);

-- samples
create index if not exists samples_timestamp_index on samples(timestamp);
create index if not exists samples_timeseries_index on samples(timeseries_id);

-- literals
create index if not exists literals_value_index on literals(value);
`)
	if err != nil {
		log.Info("Failed to initialize database", zap.String("driver", "sqlite"), zap.Error(err))
	} else {
		log.Info("Database Initialized", zap.String("driver", "sqlite"))
	}
	rows, err := db.Query("select id, value from literals")
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		var value string
		err := rows.Scan(&id, &value)
		if err != nil {
			return err
		}
		d.literalCache.Set(value, id)
	}
	d.writeQueue = d.startWriteQueue()
	return nil
}

var ErrWriteQueueIsAlreadyClosed = errors.New("write queue is already closed")

func (d *sqliteDriver) Write(req *prompb.WriteRequest) error {
	select {
	case d.writeQueue <- req:
		return nil
	default:
		return ErrWriteQueueIsAlreadyClosed
	}
}

func (d *sqliteDriver) startWriteQueue() chan<- *prompb.WriteRequest {
	log := zap.L()
	q := make(chan *prompb.WriteRequest, 100)
	go func() {
		d.waitGroup.Add(1)
		defer d.waitGroup.Done()
		log.Info("Writer thread started")
		for {
			select {
			case req, ok := <-q:
				if !ok {
					return
				}
				err := d.write(req)
				if err != nil {
					log.Error("Failed to write request", zap.Error(err))
				}
			}
		}
	}()
	return q
}

func (d *sqliteDriver) writeLiteral(literal string) (bool, error) {
	log := zap.L()
	var err error
	if _, ok := d.literalCache.Get(literal); ok {
		return false, nil
	}
	result, err := db.Exec("insert or ignore into literals (value) values (?); select last_insert_rowid()", literal)
	if err != nil {
		log.Error("Failed to append literals", zap.Error(err))
		return false, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Error("Failed to read id", zap.Error(err))
		return false, err
	}
	d.literalCache.Insert(literal, id)
	log.Info("inserted", zap.Int64("id", id), zap.String("literal", literal))
	return true, nil
}

func (d *sqliteDriver) write(req *prompb.WriteRequest) error {
	log := zap.L()
	var err error
	var result sql.Result
	numLiteralsTotal := 0
	numLiteralsInserted := int64(0)
	for _, timeseries := range req.Timeseries {
		for _, label := range timeseries.Labels {
			for _, literal := range []string{label.Name, label.Value} {
				numLiteralsTotal++
				written, err := d.writeLiteral(literal)
				if err != nil {
					return err
				}
				if written {
					numLiteralsInserted++
				}
			}
		}
	}
	log.Info("Labels inserted", zap.Int("total", numLiteralsTotal), zap.Int64("inserted", numLiteralsInserted), zap.Int("total", d.literalCache.Len()))
	numLabelsTotal := 0
	numLabelsInserted := int64(0)
	labelSQL := ""
	labelValue := make([]interface{}, 0)
	numSamplesTotal := 0
	numSamplesInserted := int64(0)
	sampleSQL := ""
	sampleValue := make([]interface{}, 0)
	for _, ts := range req.Timeseries {
		result, err := db.Exec(`insert into timeseries default values; select last_insert_rowid()`)
		if err != nil {
			log.Error("Failed to create new timeseries", zap.Error(err))
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			log.Error("Failed to read new timeseries id", zap.Error(err))
			return err
		}
		for _, label := range ts.Labels {
			labelSQL += `,(?, ?, ?)`
			nameId, _ := d.literalCache.Get(label.Name)
			valueId, _ := d.literalCache.Get(label.Value)
			labelValue = append(labelValue, id, nameId, valueId)
			numLabelsTotal++
		}
		for _, sample := range ts.Samples {
			sampleSQL += `,(?,?,?)`
			sampleValue = append(sampleValue, id, sample.Timestamp, sample.Value)
			numSamplesTotal++
		}
	}

	// Label batch insert
	labelSQL = `insert into labels (timeseries_id, name, value) values ` + labelSQL[1:]
	result, err = db.Exec(labelSQL, labelValue...)
	if err != nil {
		log.Error("Failed to write labels to database", zap.Error(err))
		return err
	}
	numLabelsInserted, err = result.RowsAffected()
	if err != nil {
		log.Error("Failed to read rows affected", zap.Error(err))
		return err
	}

	// Sample batch insert
	sampleSQL = `insert into samples (timeseries_id, timestamp, value) values ` + sampleSQL[1:]
	result, err = db.Exec(sampleSQL, sampleValue...)
	if err != nil {
		log.Error("Failed to write samples to database", zap.Error(err))
		return err
	}
	numSamplesInserted, err = result.RowsAffected()
	if err != nil {
		log.Error("Failed to read rows affected", zap.Error(err))
		return err
	}

	log.Info("Write Done",
		zap.String("driver", "sqlite"),
		zap.Int("timeseries", len(req.Timeseries)),
		zap.Int("literals-total", numLiteralsTotal),
		zap.Int64("literals-inserted", numLiteralsInserted),
		zap.Int("labels-total", numLabelsTotal),
		zap.Int64("labels-inserted", numLabelsInserted),
		zap.Int("samples-total", numSamplesTotal),
		zap.Int64("samples-inserted", numSamplesInserted))
	return nil
}

func (d *sqliteDriver) Read(req *prompb.ReadRequest, w io.Writer) error {
	return nil
}

func (d *sqliteDriver) Close() {
	close(d.writeQueue)
	_ = d.db.Close()
	d.waitGroup.Wait()
}
