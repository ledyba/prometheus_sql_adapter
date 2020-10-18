package repo

import (
	"database/sql"
	"io"

	"github.com/cornelk/hashmap"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

type sqliteDriver struct {
	literalCache hashmap.HashMap
	db *sql.DB
}

func newSqlite(db *sql.DB) Driver {
	return &sqliteDriver{
		literalCache: hashmap.HashMap{},
		db: db,
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
		err = rows.Scan(&id, &value)
		if err != nil {
			return err
		}
		d.literalCache.Set(value, id)
	}
	return err
}

func (d *sqliteDriver) Write(req *prompb.WriteRequest) error {
	log := zap.L()
	var err error
	var result sql.Result
	numLiteralsTotal := 0
	numLiteralsInserted := int64(0)
	for _, timeseries := range req.Timeseries {
		for _, label := range timeseries.Labels {
			for _, literal := range []string{label.Name, label.Value} {
				numLiteralsTotal++
				if _, ok := d.literalCache.Get(literal); ok {
					continue
				}
				result, err = db.Exec("insert or ignore into literals (value) values (?)", literal)
				if err != nil {
					log.Error("Failed to append literals", zap.Error(err))
					return err
				}
				var affected int64
				affected, err = result.RowsAffected()
				if err != nil {
					log.Error("Failed to read rows affected", zap.Error(err))
					return err
				}
				numLiteralsInserted += affected
				row := db.QueryRow(`select id from literals where rowid = last_insert_rowid()`)
				if row.Err() != nil {
					log.Error("Failed to select id from timeseries where rowid = last_insert_rowid()", zap.Error(row.Err()))
					return row.Err()
				}
				var id uint64
				if err = row.Scan(&id); err != nil {
					return err
				}
				d.literalCache.Set(id, literal)
			}
		}
	}
	log.Info("Labels inserted", zap.Int("total", numLiteralsTotal), zap.Int64("inserted", numLiteralsInserted))
	numLabelsTotal := 0
	numLabelsInserted := int64(0)
	labelSQL := ""
	labelValue := make([]interface{}, 0)
	numSamplesTotal := 0
	numSamplesInserted := int64(0)
	sampleSQL := ""
	sampleValue := make([]interface{}, 0)
	for _, ts := range req.Timeseries {
		_, err = db.Exec(`insert into timeseries default values`)
		if err != nil {
			log.Error("Failed to create new timeseries", zap.Error(err))
			return err
		}
		row := db.QueryRow(`select id from timeseries where rowid = last_insert_rowid()`)
		if row.Err() != nil {
			log.Error("Failed to select id from timeseries where rowid = last_insert_rowid()", zap.Error(row.Err()))
			return row.Err()
		}
		var id uint64
		if err = row.Scan(&id); err != nil {
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
	_, err = db.Exec(labelSQL, labelValue...)
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
	_, err = db.Exec(sampleSQL, sampleValue...)
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

}